// Favorite API
import {api} from "../request";

export const favoriteApi = {
    toggle: (params: { media_id: string }) =>
        api.post("/api/v1/media/:mediaId/favorite", {}, {params}),
    getStatus: (params: { media_id: string }) =>
        api.get("/api/v1/media/:mediaId/favorite", {params}),
};