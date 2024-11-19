# Defining rules for your code

One of the interests of AST Metrics is to allow you to check that your code respects certain rules.

For example, you can check that functions do not exceed a certain complexity.

To do this, create a `.ast-metrics.yaml` file at the root of your project and add the rules you want to check.

You can also run the `ast-metrics init` command to generate this file with default rules.

We will add a rule to check that files do not exceed a complexity of 30.

```yaml
sources: 
  - ./src

requirements:
  rules:
    cyclomatic_complexity:
      max: 30
      excludes: []
```

Now you can run the `ast-metrics analyze` command to check that your code respects this rule.

```console
ast-metrics analyze
```

If all your files respect this rule, the command will return 0. If not, you will see an error message indicating which files do not respect this rule, and the command will return an error code.

For each of the rules, you can specify exceptions. For example, if you have a very complex function but cannot simplify it, you can exclude it from the check. Note that regular expressions are used to specify the files to exclude.

```yaml
requirements:
  rules:
    cyclomatic_complexity:
      max: 30
      excludes: 
        - very_complex_file
```

## Available rules

The list of available rules is growing regularly. Here is the current list:

**Avoid coupling between classes**

This constraint checks that there is no coupling between classes. For example, it may be forbidden for a controller to use a repository.

```yaml
requirements:
  rules:
    coupling:
      forbidden:
        - from: "Controller.*"
          to: ".*Repository.*"
```

**Maximum number of lines of code per file**

```yaml
requirements:
  rules:
    ...
    lines_of_code:
      max: 1000
```

**Code maintainability**

Code maintainability is a measure of how easy the code can be maintained. It is calculated based on cyclomatic complexity, number of lines of code, number of functions, and number of classes.

It ranges from 0 to 171. Generally, > 85 is considered an acceptable value.

```yaml
requirements:
  rules:
    ...
    maintainability:
      min: 85
```

**Cyclomatic complexity**

Cyclomatic complexity is a measure of the complexity of a function. It is calculated based on the number of possible paths in a function.

```yaml
requirements:
  rules:
    ...
    cyclomatic_complexity:
      max: 30
```