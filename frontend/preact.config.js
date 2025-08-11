// eslint-disable-next-line @typescript-eslint/no-var-requires
const preactCliSvgLoader = require("preact-cli-svg-loader");

export default (config, env, helpers) => {
    config.module.rules.push({
        test: /\.(woff2?|ttf|otf|eot|mp4|mov|ogg|webm)(\?.*)?$/i,
        loader: "url-loader?limit=100000",
    });
    config.module.rules.push({
        test: /\.glsl$/,
        loader: "webpack-glsl-loader",
    });
    if (env.isProd) {
        config.devtool = false;
    }
    preactCliSvgLoader(config, helpers);
};
