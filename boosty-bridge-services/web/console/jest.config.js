module.exports = {
    injectGlobals: true,
    /** To test HTML DOM add in spec.[ts | tsx] next line:
     * @jest-environment jsdom. */
    testEnvironment: 'jsdom',
    /**
     * Test enviroment options needed to make manipulation with HTML DOM.
     * Could be extended with such fields as html, url, and userAgent.
     * To make this options visible add in spec.[ts | tsx] next line:
     * @jest-environment-options.
     */
    testEnvironmentOptions: {
        html: `
            <html lang="en">
                <head>
                    <meta name="description" content="test environment options"/>
                    <meta name="gateway-address" content="http://localhost:8089">
                <head/>
            </html>
        `,
    },
    moduleDirectories: ["node_modules", "src"],
    moduleNameMapper: {
        'app/(.*)': '<rootDir>/src/app/$1',
        '^@/(.*)$': '<rootDir>/src/$1',
        '^@static/(.*)$': '<rootDir>/src/app/static/$1',
        '\\.(css|less|sass|scss)$': 'identity-obj-proxy'
    },
    roots: ['<rootDir>'],
    transform: {
        '^.+\\.(ts|tsx)?$': 'ts-jest',
        '\\.svg': 'jest-transform-stub'
    },
    testRegex: '(/tests/.*|(\\.|/)(test|spec))\\.(tsx|ts)?$',
    moduleFileExtensions: ['ts', 'js', 'tsx', 'jsx', 'json'],
    collectCoverage: true,
    clearMocks: true,
    coverageDirectory: "coverage",
};
