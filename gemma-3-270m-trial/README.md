# Gemma 3 270M 試遊

## ログ

まずはインストール。
CUDA 12.8 と Winネイティブ Python 3.13 環境。
venv でスタート

<https://ai.google.dev/gemma/docs/core/huggingface_text_full_finetune?hl=ja> に従う。
flash-attn が入らない。
torchが見つからないと言われる。

torchはCPU版が入ってたので <https://pytorch.org/get-started/locally/> に従い GPU 対応版をインストール。

```
Python 3.13.0 (tags/v3.13.0:60403a5, Oct  7 2024, 09:38:07) [MSC v.1941 64 bit (AMD64)] on win32
Type "help", "copyright", "credits" or "license" for more information.
>>> import torch
>>> torch.__version__
'2.8.0+cu128'
```

<https://github.com/Dao-AILab/flash-attention/issues/1421#issuecomment-2575547768> に従い、以下のコマンドで再挑戦。
torch 云々では怒られないが、別のところ(おそらくコンパイル)でエラー。

    pip install flash_attn --no-build-isolation

    ... (snipped) ...

    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2039: 'is_unsigned_v': 'cutlass::platform' のメンバーではありません
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/integer_subbyte.h(235): note: 'cutlass::platform' の宣言を確認してください
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(860): note: コンパイル対象の クラス テンプレート インスタンス化 'cutlass::detail::FpBitRepresentation<uint32_t,32,8,23,cutlass::detail::NanInfEncoding::IEEE_754,true>' のリファレンスを確認してください
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(950): note: コンパイル対象の関数 テンプレート インスタンス化 'auto cutlass::detail::fp_encoding_selector<cutlass::detail::FpEncoding::E8M23>(void)' のリファレンスを確認してください
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(1211): note: コンパイル対象の クラス テンプレート インスタンス化 'cutlass::float_exmy_base<T,Derived>' のリファレンスを確認してください
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2065: 'is_unsigned_v': 定義されていない識別子です。
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint32_t,32,8,23,cutlass::detail::NanInfEncoding::IEEE_754,true>::Storage': 型の代わりに式が必要です
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2059: 構文エラー: ','
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2238: ';' の前に無効なトークンがあります。
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint8_t,8,4,3,cutlass::detail::NanInfEncoding::CANONICAL_ONLY,false>::Storage': 型の代わりに式が必要です
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint8_t,8,8,0,cutlass::detail::NanInfEncoding::CANONICAL_ONLY,false>::Storage': 型の代わりに式が必要です
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint8_t,4,2,1,cutlass::detail::NanInfEncoding::NONE,true>::Storage': 型の代わりに式が必要です
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint8_t,6,2,3,cutlass::detail::NanInfEncoding::NONE,true>::Storage': 型の代わりに式が必要です
    C:\Users\koron\AppData\Local\Temp\pip-install-pa5rwxh3\flash-attn_2c2caa2b6ce14386a750cead7169aa8b\csrc\cutlass\include\cutlass/exmy_base.h(404): error C2275: 'cutlass::detail::FpBitRepresentation<uint8_t,6,3,2,cutlass::detail::NanInfEncoding::NONE,true>::Storage': 型の代わりに式が必要です

エラーメッセージの内容から考えるに、flash-attnに含まれる cutlass が Visual C++ のコンパイラでビルドに失敗している。

最初のエラーの該当箇所 <https://github.com/NVIDIA/cutlass/blob/a49a78ffefc86a87160dfe0ccc3a3a2d1622c918/include/cutlass/exmy_base.h#L403C6-L405>

`CUTLASS_CXX17_OR_LATER` がどこで定義されるかが気になる。

定義箇所 <https://github.com/NVIDIA/cutlass/blob/a49a78ffefc86a87160dfe0ccc3a3a2d1622c918/include/cutlass/detail/helper_macros.hpp#L193-L207>

Visual Studio 2022 (Visual C++ 17) の説明 <https://learn.microsoft.com/ja-jp/cpp/preprocessor/predefined-macros?view=msvc-170>

`/std:c++17` 以降を指定してコンパイルしている、 `CUTLASS_CXX17_OR_LATER` が定義されていることは確実。

`is_unsigned_v` の定義箇所: <https://github.com/NVIDIA/cutlass/blob/a49a78ffefc86a87160dfe0ccc3a3a2d1622c918/include/cutlass/platform/platform.h#L525-L532>

ラストチャンス: <https://nowokay.hatenablog.com/entry/2024/06/05/190603> の情報を元に、
`pip install whell` と `export DISTUTILS_USE_SDK=1` してから再インストール。
本当に3時間かかったら嫌だな…w その時はいったん諦める。

メモリが64GBに張り付き、たぶんスラッシングでSSDも張り付いた。
cicc (CUDAのコンパイラ)が複数起動して、食いつぶしたらしい。
並列数を抑える必要があるだろう。

pipがninjaをデフォルトの `-j 4` で起動して、
ninjaが nvcc (NVIDIAのコンパイララッパー)を4つ起動し、
各nvccが cicc (CUDAコンパイラ) を x4 で起動して、合計16個の cicc が起動してた。
で、各 cicc が 4GBくらいのメモリを消費するので、搭載された64GBの物理メモリを食いつぶし、スロッシング発生。

`MAX_JOBS=1` で起動することでビルドは進んでいるけど… `2` で、良いかも。

[ファインチューンの結果一覧](./martian-finetune-result.html)

使ってるデータセット [bebechien/MobileGameNPC](https://huggingface.co/datasets/bebechien/MobileGameNPC) は
地球人と異星人(訛り有)との会話という想定の25本のデータセット。
今回は20本を学習用、5本を確認用という振り分けをしている。

正確さ(再現度)はいまいちだが、この程度の使い方であれば雰囲気はでている。
