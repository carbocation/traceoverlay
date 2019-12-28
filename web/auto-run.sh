CompileDaemon \
    -directory=./ \
    -command='./carbocation-tools.linux -port 9019 -manifest /Users/jamesp/data/mri/manifest_20208_3ch.tsv -project demo' \
    -build='go build -o carbocation-tools.linux' \
    -pattern='(.+\.go|.+\.c|.+\.html|.+\.js)$' \
    -graceful-kill=true

# Fetch just instance_number == 1 records for LAX 3Ch:
#cat <(head -n 1 manifest.20190929.50k.tsv) <(grep _20208_ manifest.20190929.50k.tsv | grep -i 3ch | awk -F $'\t' '$9==1 {print $0}') > manifest_20208_3ch.tsv
