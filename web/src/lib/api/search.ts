// Search API
import {api} from "../request";
import {Media} from "./media";
import {Article} from "./article";

export interface SearchResponse {
    media?: Media[];
    articles?: Article[];
}

export const searchApi = {
    search: (query: string, type?: "media" | "articles" | "all") =>
        api.get<SearchResponse>("/search", {q: query, type}),
};
