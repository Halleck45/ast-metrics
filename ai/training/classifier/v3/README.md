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

## Dataset

Prepare the dataset

```bash
go run cmd/dev/ai_dataset.go --output=<path-to-output-csv> <path-to-examples-folder>
```

If you want to work on sample data, you can generate it with:

```bash
make dev-prepare-examples
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/v3/dataset/samples.csv ./tmp/samples/
```

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

## Evaluation

@todo

## Onshot
You can do all the steps in one command with the `onshot.py` script.

```bash
python 1-labelize.py --count=10000 dataset/samples.csv
python 2-merge_dataset.py dataset/samples.csv classified_output/classified_c4.csv dataset/final_dataset.csv
python 3-train.py dataset/final_dataset.csv model.pkl features.json
python 4-predict.py dataset/samples.csv model.pkl features.json
```

```bash
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/v3/dataset/ai-service.csv /home/jflepine/workdir/globalexam/packages/services/ai-service.globalexam.cloud/app
cd ai/training/classifier/v3
python 1-labelize.py --count=50 dataset/ai-service.csv
python 2-merge_dataset.py dataset/ai-service.csv classified_output/classified_c4.csv dataset/final_dataset.csv
python 3-train.py dataset/final_dataset.csv model.pkl features.json
python 4-predict.py dataset/ai-service.csv model.pkl features.json
```