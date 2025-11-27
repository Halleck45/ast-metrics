# Metrics Overview

AST Metrics provides a comprehensive set of metrics to help you understand the quality, structure, and health of your codebase.

We believe that **code quality is not just about style**. It's about:
- **Reliability**: How likely is it to break?
- **Maintainability**: How easy is it to change?
- **Architecture**: Does the code structure match your mental model?

## Available Metrics

### 📏 Volume & Complexity
- [**Volume**](volume.md): Lines of code, logical lines, comments. The baseline for everything.
- [**Cyclomatic Complexity**](cyclomatic-complexity.md): How many paths through your code?
- [**Maintainability Index**](maintainability-index.md): A global score for code health.
- [**Risk Score**](risk.md): Complexity × Churn. Where are the bugs hiding?

### 🔗 Coupling & Cohesion
- [**Coupling & Instability**](coupling.md): How classes depend on each other.
- [**LCOM4**](lcom4.md): Do methods in a class belong together?

### 🏗️ Architecture & Team
- [**Community Detection**](community-detection.md): The natural structure of your code.
- [**Bus Factor**](bus-factor.md): Knowledge distribution and risk.
- [**Architecture Map**](architecture-map.md): Visualizing the system.
