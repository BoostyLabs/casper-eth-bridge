import { META_TAGS_CONFIG, parseMetaTag } from '@/app/internal/parseMetaTag';
import { HttpClient } from '@/private/http/client';

/**
 * ErrorUnauthorized is a custom error type which indicates that the client request has not been
 * completed because it lacks valid authentication credentials for the requested resource.
 */
export class UnauthorizedError extends Error {
    public constructor(message = 'authorization required') {
        super(message);
    }
}

/**
 * BadRequestError is a custom error type which indicates that the server cannot or
 * will not process the request due to something that is perceived to be a client error.
 */
export class BadRequestError extends Error {
    public constructor(message = 'bad request') {
        super(message);
    }
}

/**
 * NotFoundError is a custom error type which indicates that the server can't find the requested resource.
 */
export class NotFoundError extends Error {
    public constructor(message = 'not found') {
        super(message);
    }
}

/**
 * InternalError is a custom error type which indicates that the server encountered an unexpected condition
 * that prevented it from fulfilling the request.
 */
export class InternalError extends Error {
    public constructor(message = 'internal server error') {
        super(message);
    }
}

/**
 * TooManyRequestError is a custom error type which indicates the user
 * has sent too many requests in a given amount of time.
 */
export class TooManyRequestError extends Error {
    /** Error message while bad request */
    constructor(message = 'Too many requests') {
        super(message);
    };
};

/**
 * TooLargeFileError is a custom error type which indicates the user
 * has sent a big file.
 */
export class TooLargeFileError extends Error {
    /** Error message while bad request */
    constructor(message = 'Too large file') {
        super(message);
    };
};

const BAD_REQUEST_ERROR = 400;
const UNAUTORISED_ERROR = 401;
const TOO_LARGE_FILE_ERROR = 413;
const TOO_MANY_REQUESTS_ERROR = 429;
const NOT_FOUND_ERROR = 404;
const INTERNAL_ERROR = 500;

/** Exposes filtering by type options */
export class TypeOption {
    constructor(
        public value: string,
        public isSelected: boolean = false
    ) { }
}

/**
 * APIClient is base client that holds http client and error handler.
 */
export class APIClient {
    protected readonly ROOT_PATH: string = `${parseMetaTag(META_TAGS_CONFIG.GATEWAY_ADDRESS)}/api/v0`;
    protected readonly http: HttpClient = new HttpClient();
    /**
     * handles error due to response code.
     * @param response - response from server.
     *
     * @throws {@link NotFoundError}
     * This exception is thrown if the input is not a valid ISBN number.
     *
     * @throws {@link UnauthorizedError}
     * Thrown if the ISBN number is valid, but no such book exists in the catalog.
     *
     * @throws {@link InternalError}
     * Thrown if the ISBN number is valid, but no such book exists in the catalog.
     *
    * @throws {@link TOO_LARGE_FILE_ERROR}
     * Thrown if the ISBN number is valid, but no such book exists in the catalog.
     *
     * @private
     */
    /* eslint-disable */
    protected async handleError(response: Response): Promise<void> {
        const body = await response.json();
        switch (response.status) {
            case BAD_REQUEST_ERROR:
                throw new BadRequestError(body.error);
            case TOO_LARGE_FILE_ERROR:
                throw new TooLargeFileError(body.error);
            case NOT_FOUND_ERROR:
                throw new NotFoundError(body.error);
            case UNAUTORISED_ERROR:
                throw new UnauthorizedError(body.error);
            case TOO_MANY_REQUESTS_ERROR:
                throw new TooManyRequestError(body.error);
            case INTERNAL_ERROR:
            default:
                throw new InternalError(body.error);
        }
    }
}
