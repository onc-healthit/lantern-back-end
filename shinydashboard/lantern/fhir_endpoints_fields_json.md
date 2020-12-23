# JSON Exportable File Format

### Fields
&nbsp;

* **url:** Service base URL of the endpoint  
* **api_information_source_name:** Organization names associated with the endpoint reported by the list source  
* **created_at:** Timestamp of the endpoint's record creation in the Lantern database  
* **list_source:** Name or URL of the list source that included the endpoint  
* **certified_api_developer_name:** Name of the API developer associated with the endpoint  
* **operation:** See [below](#operation).  
&nbsp;

### Operation
\
The operation field is an array where an element is an instance of the information Lantern received when querying the given endpoint.  
&nbsp;

* **http_response:** HTTP response received from querying the endpoint's metadata url  
* **http_response_time_second:** HTTP response time of the endpoint  
* **errors:** Errors receieved from querying the endpoint  
* **fhir_version:** FHIR version that is pulled from the endpoint's capability statement  
* **tls_verison:** Transport Layer Security (TLS) version of the endpoint  
* **mime_types:** MIME types supported by this endpoint  
* **updated:** Timestamp of when this endpoint record was updated  
* **supported_resources:** All of the FHIR resources this endpoint supports  
* **smart_http_response:** HTTP response received from querying the endpoint's SMART url  
* **smart_response:** See [below](#smart-response).  
&nbsp;

### SMART Response
\
The SMART Response field is the value received from querying the given URL's `/.well-known/configuration` endpoint. More information about this and SMART on FHIR can be found [here](http://www.hl7.org/fhir/smart-app-launch/conformance/index.html).  
&nbsp;

* **authorization_endpoint:** The URL to the OAuth2 authorization endpoint  
* **token_endpoint:** The URL to the OAuth2 token endpoint  
* **token_endpoint_auth_methods:** An array of client authentication methods supported by the given token endpoint  
* **registration_endpoint:** The URL to the OAuth2 dynamic registration endpoint for this FHIR server  
* **scopes_supported:** An array of scopes a client may request  
* **response_types_supported:** An array of OAuth2 `response_type` values that are supported  
* **management_endpoint:** The URL where an end-user can view which applications currently have access to data and can make adjustments to these access rights  
* **introspection_endpoint:** The URL to a server’s introspection endpoint that can be used to validate a token  
* **revocation_endpoint:** The URL to a server’s revoke endpoint that can be used to revoke a token  
* **capabilities:** An array representing SMART capabilities that the server supports  