// Like API
import {api} from "../request";

export const likeApi = {
    toggle: (params: { media_id: string }) =>
        api.post("/api/v1/media/:mediaId/like", {}, {params}),
    getStatus: (params: { media_id: string }) =>
        api.get("/api/v1/media/:mediaId/like", {params}),
};