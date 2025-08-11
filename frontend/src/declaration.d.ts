declare module "*.css" {
    const mapping: Record<string, string>;
    export default mapping;
}

declare module "*.svg" {
    const content: FunctionComponent;
    export default content;
}

declare module "*.glsl" {
    const content: string;
    export default content;
}
