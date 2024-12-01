set -e

# number of packages to download
PACKAGES_COUNT=$1
if [ -z "$PACKAGES_COUNT" ]; then
    PACKAGES_COUNT=100
fi

# keep always the same workdir, to avoid download time for each package
workdir="build/monkey-test"
echo "Working in $workdir"
if [ -z "$workdir" ]; then
    echo "Workdir not found"
    exit 1
fi
if [ ! -d "$workdir" ]; then
    echo "Workdir not found, creating it"
    mkdir -p $workdir
fi

# cleanup reports
rm -f ast-metrics-report.json


# sort TOP packages randomly
url="https://packagist.org/explore/popular.json?per_page=100"
# shuffle 100 packages
packages=$(curl -s $url | jq -r '.packages[].name' | shuf)
# take only $PACKAGES_COUNT packages
packages=$(echo "$packages" | head -n $PACKAGES_COUNT)

echo "Downloading $PACKAGES_COUNT packages"
for package in $packages; 
do 
    echo "  Downloading $package"
	
    # generate md5 checksum for destination
    checksum=$(echo $package | md5sum | awk '{ print $1 }')
    destination="$workdir/$checksum"

    if [ ! -d $destination ]; then
        repository=$(curl -s https://packagist.org/packages/$package.json | jq -r '.package.repository')
        zipUrl="$repository/archive/refs/heads/master.zip"
        echo "    Downloading $zipUrl to $destination"
        curl -s -L -o $destination.zip $zipUrl
        # if zip contains HTML, like "Just a moment...", then skip
        if grep -q "<html" $destination.zip; then
            echo "  Skipping $package because it contains HTML (probably rate limited)"
            continue
        fi

        # if contains 404, then skip
        if grep -q "404" $destination.zip; then
            echo "  Skipping $package because it contains 404"
            continue
        fi

        unzip $destination.zip -d $destination > /dev/null
        rm $destination.zip
    else
        echo "  Skipping $package because it already exists"
    fi

done

echo "Analyzing $workdir"
time go run . analyze --ci $workdir

# Ensure that report is generated
if [ ! -f ast-metrics-report.json ]; then
    echo "Report not generated"
    exit 1
else 
    echo "Report generated"
fi


# Count number of analyzed files
# | **PHP** | 122.0 K | 🟢 112 | 1.21 | 12 |
line=$(cat build/report.md |grep '**PHP**'|head -n 1)
separator="|"
linesOfCode=$(echo $line | awk -F "$separator" '{print $3}')
echo "Analyzed $linesOfCode lines of code"


echo "Done"