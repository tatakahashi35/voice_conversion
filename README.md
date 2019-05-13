# voice_conversion
パラメータを修正することであらゆる人物の音声にリアルタイムに変換することを目標

## 特徴
- リアルタイムで変換する
- 深層学習を使っていない
- 変換後の音声は入力話者の音声に強く影響を受ける

## アルゴリズム
1. マイクから音声を取得する
2. mcepとF0を計算する
3. それぞれ変換する
4. MLSAフィルタにより音声合成する
5. マイクから出力する

## 今後
- 深層学習による手法を使うなどして、音質の改善をする
- 対象をある程度絞ることで、対象話者の音声との距離を近づける

## その他
リアルタイム性の維持、学習データ量などは考慮する必要がある
