#!/bin/bash

cd ../shinydashboard/lantern
echo "Installing/checking for lintr package..."
echo 'install.packages("lintr", dependencies = TRUE, repos="http://cran.rstudio.com/")' | R --save
echo "Running R lintr..."
LINTR=$(Rscript -e "lintr::lint_dir(linters = lintr::with_defaults(object_usage_linter=NULL, closed_curly_linter = NULL, open_curly_linter = NULL, line_length_linter = NULL, object_name_linter = NULL))")
if [[ ! -z "${LINTR[0]}" ]]; then
    for i in "${LINTR[@]}"
    do
        echo "$i"
    done
    exit 1
fi