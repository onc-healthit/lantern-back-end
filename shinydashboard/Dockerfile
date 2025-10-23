FROM --platform=linux/x86_64 rocker/shiny-verse:4.1.3
ARG cert_dir
ARG LANTERN_DBNAME
ARG LANTERN_DBPORT
ARG LANTERN_DBHOST
ARG LANTERN_DBUSER
ARG LANTERN_DBPASSWORD
ARG LANTERN_DBSSLMODE
ARG LANTERN_BANNER_TEXT
ARG LANTERN_CAPQUERY_QRYINTVL

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
# Install system dependencies required to compile packages like 'curl' and spatial libs (terra)
RUN apt-get update && \
		DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
			build-essential \
			libcurl4-openssl-dev \
			pkg-config \
			curl \
			libssl-dev \
			libxml2-dev \
			ca-certificates \
			git \
			gdal-bin \
			libgdal-dev \
			proj-bin \
			libproj-dev \
			libgeos-dev \
			libudunits2-dev \
			cmake \
			libharfbuzz-dev \
			libfribidi-dev \
			libfreetype6-dev \
			libpng-dev \
			libcairo2-dev \
			libglib2.0-dev \
			libfontconfig1-dev && \
		rm -rf /var/lib/apt/lists/*

RUN R -e 'renv::restore()'