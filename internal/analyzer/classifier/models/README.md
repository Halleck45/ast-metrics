# Modèles embarqués

Ce répertoire contient les modèles de classification embarqués dans le binaire via `embed`.

## Structure

Les modèles doivent être organisés par langage de programmation :
```
models/
  {lang}/
    model.json.gz
```

Par exemple :
```
models/
  php/
    model.json.gz
  java/
    model.json.gz
```

## Ajout d'un nouveau modèle

1. Créer un répertoire pour le langage : `models/{lang}/`
2. Copier le fichier `model.json.gz` depuis `ai/training/classifier/build/{lang}/model.json.gz`
3. Le modèle sera automatiquement embarqué dans le binaire lors du build

## Chargement

Le code charge automatiquement les modèles embarqués en priorité, puis fait un fallback sur le système de fichiers si le modèle embarqué n'existe pas.


