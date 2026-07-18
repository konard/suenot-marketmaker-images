#include <png.h>
#include <stdio.h>
#include <stdlib.h>

#define TARGET_WIDTH 1664
#define TARGET_HEIGHT 936

static void fail(const char *message) {
    fprintf(stderr, "%s\n", message);
    exit(EXIT_FAILURE);
}

int main(int argc, char **argv) {
    if (argc != 3) fail("usage: resize_png INPUT OUTPUT");
    FILE *input = fopen(argv[1], "rb");
    if (!input) fail("cannot open input");
    png_image image = {0};
    image.version = PNG_IMAGE_VERSION;
    if (!png_image_begin_read_from_stdio(&image, input)) fail(image.message);
    image.format = PNG_FORMAT_RGBA;
    png_bytep source = malloc(PNG_IMAGE_SIZE(image));
    if (!source || !png_image_finish_read(&image, NULL, source, 0, NULL)) fail(image.message);
    fclose(input);

    png_bytep target = malloc((size_t)TARGET_WIDTH * TARGET_HEIGHT * 4);
    if (!target) fail("cannot allocate output");
    for (int y = 0; y < TARGET_HEIGHT; ++y) {
        double sy = ((y + 0.5) * image.height / TARGET_HEIGHT) - 0.5;
        int y0 = (int)sy;
        double fy = sy - y0;
        if (y0 < 0) { y0 = 0; fy = 0; }
        int y1 = y0 + 1 < (int)image.height ? y0 + 1 : y0;
        for (int x = 0; x < TARGET_WIDTH; ++x) {
            double sx = ((x + 0.5) * image.width / TARGET_WIDTH) - 0.5;
            int x0 = (int)sx;
            double fx = sx - x0;
            if (x0 < 0) { x0 = 0; fx = 0; }
            int x1 = x0 + 1 < (int)image.width ? x0 + 1 : x0;
            for (int channel = 0; channel < 4; ++channel) {
                double top = source[((size_t)y0 * image.width + x0) * 4 + channel] * (1 - fx)
                           + source[((size_t)y0 * image.width + x1) * 4 + channel] * fx;
                double bottom = source[((size_t)y1 * image.width + x0) * 4 + channel] * (1 - fx)
                              + source[((size_t)y1 * image.width + x1) * 4 + channel] * fx;
                target[((size_t)y * TARGET_WIDTH + x) * 4 + channel] = (png_byte)(top * (1 - fy) + bottom * fy + 0.5);
            }
        }
    }

    FILE *output = fopen(argv[2], "wb");
    if (!output) fail("cannot open output");
    png_image result = {0};
    result.version = PNG_IMAGE_VERSION;
    result.width = TARGET_WIDTH;
    result.height = TARGET_HEIGHT;
    result.format = PNG_FORMAT_RGBA;
    if (!png_image_write_to_stdio(&result, output, 0, target, 0, NULL)) fail(result.message);
    fclose(output);
    free(source);
    free(target);
    return EXIT_SUCCESS;
}
