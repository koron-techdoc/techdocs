package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	"github.com/k0kubun/pp/v3"

	// Necessary to enable S3(s3://) scheme
	_ "github.com/apache/iceberg-go/io/gocloud"
)

var (
	optARN      string
	optDownload bool
	optOutdir   string
)

func main() {
	flag.StringVar(&optARN, "arn", "", `ARN for S3 Tables bucket`)
	flag.BoolVar(&optDownload, "download", false, `Download data files for tables`)
	flag.StringVar(&optOutdir, "outdir", ".", `Output dir for downloaded data files`)
	flag.Parse()

	pp.Default.SetOmitEmpty(true)
	pp.Default.SetColoringEnabled(false)

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	// Parse the ARN to extract the service name and region.
	parts := strings.SplitN(optARN, ":", 5)
	if len(parts) < 4 {
		return fmt.Errorf("too short ARN: want=4 got=%d", len(parts))
	}
	service, region := parts[2], parts[3]

	// Load AWS default configuration.
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return err
	}

	// Create a catalog to access S3 Tables.
	cat, err := rest.NewCatalog(
		ctx,
		"Arbitrary catalog name",
		fmt.Sprintf("https://s3tables.%s.amazonaws.com/iceberg", region),
		rest.WithWarehouseLocation(optARN),
		rest.WithSigV4RegionSvc(region, service),
		rest.WithAwsConfig(cfg),
	)
	if err != nil {
		return err
	}

	// List namespaces
	namespaces, err := cat.ListNamespaces(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("namespaces=%+v\n", namespaces)

	// Prepare S3 client to access the head of data file.
	client := s3.NewFromConfig(cfg)

	// Retrieve data files for all tables from each namespace.
	for _, ns := range namespaces {
		for id, err := range cat.ListTables(ctx, ns) {
			if err != nil {
				return err
			}
			fmt.Println()
			fmt.Printf("%s\n", id)
			table, err := cat.LoadTable(ctx, id)
			if err != nil {
				return err
			}
			scan := table.Scan()
			tasks, err := scan.PlanFiles(ctx)
			if err != nil {
				return err
			}
			for i, task := range tasks {
				// Retrieve the header of a data file from a path using the AWS S3 SDK.
				path := task.File.FilePath()
				fmt.Printf("#%d FilePath=%s\n", i, path)
				if optDownload {
					_, err := getBody(ctx, client, path, optOutdir)
					if err != nil {
						return err
					}
				} else {
					headOut, err := getHead(ctx, client, path)
					if err != nil {
						return err
					}
					if headOut != nil {
						pp.Println(headOut)
					}
				}
			}
		}
	}

	return nil
}

// getHead retrieves the S3 header object using the Iceberg's data file path
// format.
func getHead(ctx context.Context, client *s3.Client, path string) (*s3.HeadObjectOutput, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	bucket, key := u.Host, strings.TrimLeft(u.Path, "/")
	fmt.Printf("getHead\n    bucket=%s\n    key=%s\n", bucket, key)

	headOut, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: new(bucket),
		Key:    new(key),
	})
	if err != nil {
		return nil, err
	}
	headOut.ResultMetadata = middleware.Metadata{}
	return headOut, nil
}

func getBody(ctx context.Context, client *s3.Client, s3url, outdir string) (string, error) {
	// Parse S3 URL
	u, err := url.Parse(s3url)
	if err != nil {
		return "", err
	}
	bucket, key := u.Host, strings.TrimLeft(u.Path, "/")
	name := path.Base(key)
	if name == "." || name == "/" {
		return "", fmt.Errorf("invalid name: %q", name)
	}
	localName := filepath.Join(outdir, name)

	fmt.Printf("getBody    bucket=%s\n    key=%s\n    name=%s\n    \nlocalName=%s\n", bucket, key, name, localName)

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: new(bucket),
		Key:    new(key),
	})
	if err != nil {
		return "", err
	}
	defer out.Body.Close()

	if err := os.MkdirAll(outdir, 0755); err != nil {
		return "", err
	}
	localFile, err := os.Create(localName)
	if err != nil {
		return "", err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, out.Body); err != nil {
		return "", err
	}

	return localName, nil
}
