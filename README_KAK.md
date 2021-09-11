### 证明参数

kak的主链支持了8G的扇区，也降低了P盘时HASH运算的层数，所以证明参数要重新生成。

下载：

~~~
git clone https://github.com/filecoin-project/rust-fil-proofs.git
git clone https://gitee.com/ka-klab_1/kak-code.git
~~~

对官方代码做了修改，代码提交在了kak-code的代码仓库，需要用用kak-code/extern/filecoin-proofs-6.1.0和kak-code/extern/storage-proofs-porep-6.1.0覆盖掉rust-fil-proofs里的代码。

编译：

~~~
cd filecoin-project/rust-fil-proofs
cargo build --release --all
~~~

就会生成相应的工具（paramcache、parampublish、fakeipfsadd），工具用来生成证明参数。

生成：

~~~
生成证明参数：工具提供了界面，可以选择不同的P盘大小。
./target/release/paramcache
发布证明参数，根据生成的证明参数，生成配置文件parameters.json：
./target/release/parampublish --ipfs-bin=./target/release/fakeipfsadd -a
~~~

把新生成的parameters.json拷贝到kak-code的代码路径。并确定/var/tmp/filecoin-proof-parameters目录保存了新的证明参数。

普通用户的证明参数不用自己生成，使用kak-code生成的就好，后期kak需要搭建一台ipfs的服务器，存放kak-code的官方证明参数。
