/**
 * @jest-environment jsdom
 * @jest-environment-options
 */

import { parseMetaTag } from '@app/internal/parseMetaTag';

describe('parses HTML meta tag and returns content', () => {
    const DOMAIN_ADDRESS: string = 'http://localhost:8089';
    const DESCRIPTION_CONTENT: string = 'test environment options';

    it('should return content', () => {
        expect(parseMetaTag('gateway-address')).toBe(DOMAIN_ADDRESS);
        expect(parseMetaTag('description')).toBe(DESCRIPTION_CONTENT);
    });

    it('should not return content if meta tag does not exist', () => {
        expect(parseMetaTag('gateway')).not.toBe(DOMAIN_ADDRESS);
    });
});
