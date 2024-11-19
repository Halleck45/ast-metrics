replaceCodeBlock = async () => {
    const url = 'https://api.github.com/repos/Halleck45/ast-metrics/tags';
    const tags = await fetch(url).then(_ => _.json());
    const version = tags[0]['name'];

    // detect all "--latest_version--" tag in .md-content, and replace it with version
    const article = document.querySelector('.md-content article')
    if (!article) {
        return;
    }

    // replace --latest_version--, but block by block to avoid to lose DOM events
    let blocks = article.querySelectorAll('pre, code, p');
    blocks.forEach((block) => {
        console.log(block.innerHTML);
        if (block.innerHTML.includes('--latest_version--')) {
            block.innerHTML = block.innerHTML.replace(/--latest_version--/g, version);
        }
    });
    
  }
  
  replaceCodeBlock()