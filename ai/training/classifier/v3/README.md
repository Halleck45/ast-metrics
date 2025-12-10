# Modelisation

We wants to create a model to classify code snippets according to their architectural role.

## Installation


Install dependencies

```bash
pip install fasttext numpy pandas scikit-learn sentence-transformers  llama-cpp-python
```

Download GGUF models

```bash
mkdir -p models
curl -L "https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf" -o ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

I've tested many models from HuggingFace, and the Mistral 7B instruct model seems to work well for this task.


## Oneshot (if you don't want to go through all the steps)

These is a oneshot example of the whole process:

```bash
cd ai/training/classifier/v3
bash train_pipeline.bash --language=php --source=../../../../samples/php
```

And update the labels:

```bash
python ai/training/classifier/v3/6-generate-labels.py \
  ai/training/classifier/v3/labels/c4.csv \
  internal/analyzer/classifier/labels.go
```

## Dataset

Prepare the dataset

```bash
go run cmd/dev/ai_dataset.go --output=<path-to-output-csv> <path-to-examples-folder>
```

If you want to work on sample data, you can generate it with:

```bash
make dev-prepare-examples
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/v3/dataset/samples.csv ./samples/
```

The dataset I use locally has:

+ 4,881,687 lines of PHP code
+ 13,053,808 lines of Go code
+ 7,398,526 lines of Python code
+ 4,987,061 lines of Rust code

Around 30 millions lines of code.

## Labeling

You need to label the dataset. You can use a LLM to help you with that.

```bash
python 1-labelize.py --count=<number-of-lines-to-label> <path-to-csv-dataset>
```

For example:

```bash
python 1-labelize.py --count=10000 dataset/samples.csv
```

Be patient, this can take a while depending on the number of lines to label and your GPU performance.

> If your GPU is not powerful enough, you can use a remote API (OpenAI). See the `1-labelize_remote.py` script for that.

## Merge labels

Once you have labeled the dataset, you need to merge the labels with the original dataset.

```bash
python 2-merge_dataset.py <path-to-csv-dataset> <path-to-labeled-csv> <path-to-output-csv>
```

For example:

```bash
python 2-merge_dataset.py dataset/samples.csv classified_output/classified_c4.csv dataset/final_dataset.csv
``` 

## Training

Train the model

```bash
python 3-train.py <path-to-final-dataset-csv> <path-to-output-model> <path-to-feature>
```

## Prediction

Make predictions

```bash
python 4-predict.py <path-to-csv-you-want-to-test> <path-to-model> <path-to-feature>
```

## Group the files

You can group the files by their predicted label

```bash
mkdir -p build
mv features.json build/
mv model.pkl build/model_role_classifier.pkl
```

## Evaluation

