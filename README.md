# hugo-algolia-updater with Chinese segmentation support

you need config your hugo site to generate index.json:

```yaml
outputs:
  home:
    - HTML
    - RSS
    ## we need JSON here because we want hugo generate `public/index.json`
    - JSON
 ```

## 快速开始

### 最新更新

1. 完全使用go进行分词，摒弃node.js分词
2. 加入sego分词，使用双分词优化质量，同时优化分词速度
3. 加入缓存机制，每次通过md5比对文件，只对有变化的文件分词
4. 支持上传索引时使用http代理
5. 支持使用自定义分词字典，自定义停用词

### 下载
[release](https://github.com/ttys3/hugo-algolia-updater/releases)页面内下载压缩包

解压压缩包到`hugo project`根目录中执行.

## credits

this project was originally forked from <https://github.com/naah69/Hugo-Algolia-Chinese-Builder>