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

ADD /download/plumber.R /home/plumber/plumber.R
ADD /download/restendpoints.R /home/plumber/restendpoints.R

# to launch your docker container
CMD ["/home/plumber/plumber.R"]
