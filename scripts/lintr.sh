#!/bin/sh
cd ../shinydashboard/lantern
echo "Running R lintr"
LINTR=$(Rscript -e "lintr::lint_dir(linters = lintr::with_defaults(object_usage_linter=NULL, closed_curly_linter = NULL, open_curly_linter = NULL, line_length_linter = NULL, object_name_linter = NULL))")
if [[ ! -z "${LINTR[0]}" ]]; then
    for i in "${LINTR[@]}"
    do
        echo "$i"
    done
    exit 1
fi