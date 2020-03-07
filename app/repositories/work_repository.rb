# frozen_string_literal: true

class WorkRepository < ApplicationRepository
  def resent(after: "")
    query = load_query "db/work_list/get_works.graphql"
    result = Annict::GraphQL::InternalClient.new(viewer: viewer).execute(query, variables: { after: after })
    data = result.to_h.dig("data", "works")
    works = data["nodes"].select(&:present?).map do |node|
      {
        id: node["annictId"],
        title: node["title"],
        title_kana: node["titleKana"],
        title_en: node["titleEn"],
        media_text: media_text(node["media"]),
        started_on: node["startedOn"],
        syobocal_tid: node["syobocalTid"],
        syobocal_url: node["syobocalUrl"],
        mal_anime_id: node["malAnimeId"],
        mal_anime_url: node["malAnimeUrl"],
        image_url: node.dig("image", "internalUrl"),
        watchers_count: node["watchersCount"],
        disappeared_at: node["disappearedAt"]
      }
    end

    DB::WorkList::WorkConnectionEntity.new(
      works: works,
      has_next_page: data.dig("pageInfo", "hasNextPage"),
      start_cursor: data.dig("pageInfo", "startCursor")
    )
  end

  private

  def media_text(media)
    Work.media.find_value(media.downcase).text
  end
end
