// Search API
import {api} from "../request";
import {Media} from "./media";
import {Content} from "./content";

export interface SearchResponse {
    media?: Media[];
    content?: Content[];
}

export const searchApi = {
    search: (query: string, type?: "media" | "content" | "all") =>
        api.get<SearchResponse>("/search", {q: query, type}),
};
