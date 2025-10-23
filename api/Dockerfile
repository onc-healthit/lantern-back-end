FROM rstudio/plumber

ARG cert_dir
ARG LANTERN_DBNAME
ARG LANTERN_DBPORT
ARG LANTERN_DBHOST
ARG LANTERN_DBUSER
ARG LANTERN_DBPASSWORD
ARG LANTERN_DBSSLMODE

COPY ${cert_dir}/ /usr/local/share/ca-certificates
RUN update-ca-certificates

# list all the needed packages here in the same fashion
RUN R -e "install.packages('broom')"
RUN R -e "install.packages('RPostgres')"
RUN R -e "install.packages('dbplyr')"
RUN R -e "install.packages('tidyverse')"
RUN R -e "install.packages('here')"
RUN R -e "install.packages('cachem')"

RUN apt update -qq \
    && apt install --yes --no-install-recommends \
        r-cran-rpostgresql  libpq-dev

WORKDIR /home/plumber

ENV RENV_VERSION 0.13.2
RUN R -e "install.packages('remotes', repos = c(CRAN = 'https://cloud.r-project.org'))"
RUN R -e "remotes::install_github('rstudio/renv@${RENV_VERSION}')"

ADD /download/plumber.R /home/plumber/plumber.R
ADD /download/restendpoints.R /home/plumber/download/restendpoints.R
ADD /download/downloadsmodule.R /home/plumber/download/downloadsmodule.R
ADD /common/db_connection.R /home/plumber/common/db_connection.R
ADD /common/endpoints.R /home/plumber/common/endpoints.R
ADD /common/http_codes.csv /home/plumber/download/http_codes.csv

# Environment variables are passed to R via .Renviron file the container
RUN env | grep LANTERN > /home/plumber/.Renviron

# to launch your docker container
CMD ["/home/plumber/plumber.R"]
