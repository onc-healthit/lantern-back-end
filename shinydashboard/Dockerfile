FROM rocker/shiny-verse:3.6.3
ARG cert_dir

COPY ${cert_dir}/ /usr/local/share/ca-certificates
RUN update-ca-certificates

ENV RENV_VERSION 0.9.3
RUN R -e "install.packages('remotes', repos = c(CRAN = 'https://cloud.r-project.org'))"
RUN R -e "remotes::install_github('rstudio/renv@${RENV_VERSION}')"

WORKDIR /.
COPY renv.lock renv.lock
RUN R -e 'renv::restore()'