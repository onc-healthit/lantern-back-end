FROM rocker/shiny-verse:4.1.3
ARG cert_dir
ARG LANTERN_DBNAME
ARG LANTERN_DBPORT
ARG LANTERN_DBHOST
ARG LANTERN_DBUSER
ARG LANTERN_DBPASSWORD
ARG LANTERN_DBSSLMODE
ARG LANTERN_BANNER_TEXT
ARG LANTERN_CAPQUERY_QRYINTVL
ARG LANTERN_RECAPTCHA_SITE_KEY
ARG LANTERN_RECAPTCHA_SECRET

COPY ${cert_dir}/ /usr/local/share/ca-certificates
RUN update-ca-certificates

ENV RENV_VERSION 0.13.2
RUN R -e "install.packages('remotes', repos = c(CRAN = 'https://cloud.r-project.org'))"
RUN R -e "remotes::install_github('rstudio/renv@${RENV_VERSION}')"

# Environment variables are passed to R via .Renviron file the container
RUN env | grep LANTERN > /home/shiny/.Renviron && chown shiny.shiny /home/shiny/.Renviron
COPY shiny-server.conf  /etc/shiny-server/shiny-server.conf

WORKDIR /.
COPY renv.lock renv.lock
RUN R -e 'renv::restore()'