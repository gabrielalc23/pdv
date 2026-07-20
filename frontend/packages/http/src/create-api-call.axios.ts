import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import type { input, output, ZodSafeParseResult, ZodType } from "zod";
import { InvalidApiResponseError } from "@pdv/errors";
import { resolveRequestLocation } from "@pdv/utils";
import type { ApiRequestLocation } from "@pdv/utils";
import type { HttpMethod } from "./http-method.axios";
import { instance } from "./instance.axios";
import { instanceWithoutInterceptors } from "./instance-without-interceptors.axios";

export interface CreateApiCallOptions<
  TRequestSchema extends ZodType,
  TResponseSchema extends ZodType,
> {
  type?: "private" | "public";
  method: HttpMethod;
  path: string;
  requestLocation?: ApiRequestLocation;
  requestSchema: TRequestSchema;
  responseSchema: TResponseSchema;
}

export function createApiCall<
  TRequestSchema extends ZodType,
  TResponseSchema extends ZodType,
>({
  type = "private",
  method,
  path,
  requestLocation,
  requestSchema,
  responseSchema,
}: CreateApiCallOptions<TRequestSchema, TResponseSchema>): (
  requestData: input<TRequestSchema>,
) => Promise<output<TResponseSchema>> {
  return async (
    requestData: input<TRequestSchema>,
  ): Promise<output<TResponseSchema>> => {
    const parsedRequest: output<TRequestSchema> =
      requestSchema.parse(requestData);

    const resolvedRequestLocation: ApiRequestLocation = resolveRequestLocation(
      method,
      requestLocation,
    );

    const requestConfig: AxiosRequestConfig = {
      method,
      url: path,
    };

    if (resolvedRequestLocation === "params") {
      requestConfig.params = parsedRequest;
    }

    if (resolvedRequestLocation === "data") {
      requestConfig.data = parsedRequest;
    }

    const axiosInstance: AxiosInstance =
      type === "private" ? instance : instanceWithoutInterceptors;

    const response: AxiosResponse<unknown> =
      await axiosInstance.request<unknown>(requestConfig);

    const parsedResponse: ZodSafeParseResult<output<TResponseSchema>> =
      responseSchema.safeParse(response.data);

    if (!parsedResponse.success) {
      throw new InvalidApiResponseError(path, parsedResponse.error);
    }

    return parsedResponse.data;
  };
}
