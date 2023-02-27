/**
 * HttpClient is a custom wrapper around fetch api.
 * Exposes get, post and delete methods for JSON.
 */
export class HttpClient {
    /**
     * Performs POST http request with JSON body.
     * @param path
     * @param body serialized JSON
     */
    public async post(path: string, body?: string): Promise<Response> {
        return await this.do('POST', path, body);
    };

    /**
     * Performs PATCH http request with JSON body.
     * @param path
     * @param body serialized JSON
     */
    public async patch(path: string, body?: string): Promise<Response> {
        return await this.do('PATCH', path, body);
    };

    /**
     * Performs PUT http request with JSON body.
     * @param path
     * @param body serialized JSON
     * @param _auth indicates if authentication is needed
     */
    public async put(path: string, body?: string, _auth = true): Promise<Response> {
        return await this.do('PUT', path, body);
    };

    /**
     * Performs GET http request.
     * @param path
     * @param _auth indicates if authentication is needed
     */
    public async get(path: string, body?: string, _auth = true): Promise<Response> {
        return await this.do('GET', path);
    };

    /**
     * Performs DELETE http request.
     * @param path
     * @param _auth indicates if authentication is needed
     */
    /** TODO: DELETE method will be reworked after back-end remarks.
     * Right now needs body here. */
    public async delete(path: string, body?: string, _auth = true): Promise<Response> {
        return await this.do('DELETE', path, body);
    };

    /**
     * do sends an HTTP request and returns an HTTP response as configured on the client.
     * @param method holds http method type
     * @param path
     * @param body serialized JSON
     */
    private async do(method: string, path: string, body?: string): Promise<Response> {
        const request: RequestInit = {
            method: method,
            body: body,
        };

        request.headers = {
            'Content-Type': 'application/json',
        };

        return await fetch(path, request);
    };
};
