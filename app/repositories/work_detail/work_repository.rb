# frozen_string_literal: true

module WorkDetail
  class WorkRepository < ApplicationRepository
    def fetch(work_id:)
      result = graphql_client.execute(query, variables: { annictId: work_id.to_i })
      data = result.to_h.dig("data", "works")
      node = data["nodes"].first

      work = {
        id: node["annictId"],
        title: node["title"],
        title_en: node["title_en"],
        title_kana: node["titleKana"],
        title_alter: node["titleAlter"],
        title_alter_en: node["titleAlterEn"],
        media: node["media"].downcase,
        season_year: node["seasonYear"],
        season_type: node["seasonType"]&.downcase,
        season_slug: node["seasonSlug"],
        started_on: node["startedOn"] ? Date.parse(node["startedOn"]) : nil,
        episodes_count: node["episodesCount"],
        watchers_count: node["watchersCount"],
        satisfaction_rate: node["satisfactionRate"],
        ratings_count: node["ratingsCount"],
        work_records_with_body_count: node["workRecordsWithBodyCount"],
        official_site_url: node["officialSiteUrl"],
        official_site_url_en: node["officialSiteUrlEn"],
        wikipedia_url: node["wikipediaUrl"],
        wikipedia_url_en: node["wikipediaUrlEn"],
        twitter_username: node["twitterUsername"],
        twitter_hashtag: node["twitterHashtag"],
        syobocal_tid: node["syobocalTid"],
        mal_anime_id: node["malAnimeId"],
        is_no_episodes: node["isNoEpisodes"],
        synopsis_html: node["synopsisHtml"],
        synopsis_en_html: node["synopsis_en_html"],
        synopsis_source: node["synopsisSource"],
        synopsis_source_en: node["synopsisSourceEn"],
        copyright: node["copyright"],
        image_url_1x: node.dig("image", "internalUrl_1x"),
        image_url_2x: node.dig("image", "internalUrl_2x"),
        trailers: [],
        casts: [],
        staffs: [],
        episodes: [],
        programs: [],
        work_records: [],
        series_list: []
      }

      node["trailers"]["nodes"].map do |node|
        work[:trailers] << {
          title: node["title"],
          url: node["url"],
          image_url: node["internalImageUrl"]
        }
      end

      node["casts"]["nodes"].map do |node|
        work[:casts] << {
          accurate_name: node["accurateName"],
          accurate_name_en: node["accurateNameEn"],
          character: {
            id: node.dig("character", "annictId"),
            name: node.dig("character", "name"),
            name_en: node.dig("character", "nameEn")
          },
          person: {
            id: node.dig("person", "annictId")
          }
        }
      end

      node["staffs"]["nodes"].map do |node|
        work[:staffs] << {
          accurate_name: node["accurateName"],
          accurate_name_en: node["accurateNameEn"],
          role: node["role"],
          role_en: node["roleEn"],
          resource: {
            typename: node.dig("resource", "__typename"),
            id: node.dig("resource", "annictId")
          }
        }
      end

      node["episodes"]["nodes"].map do |node|
        work[:episodes] << {
          id: node["annictId"],
          number_text: node["numberText"],
          title: node["title"],
          title_en: node["titleEn"]
        }
      end

      node["programs"]["nodes"].map do |node|
        work[:programs] << {
          vod_title_code: node["vodTitleCode"],
          vod_title_name: node["vodTitleName"],
          vod_title_url: node["vodTitleUrl"],
          channel: {
            id: node.dig("channel", "annictId"),
            name: node.dig("channel", "name")
          }
        }
      end

      node["workRecords"]["nodes"].map do |node|
        work[:work_records] << {
          id: node["annictId"],
          rating_animation_state: node["ratingAnimationState"]&.downcase,
          rating_music_state: node["ratingMusicState"]&.downcase,
          rating_story_state: node["ratingStoryState"]&.downcase,
          rating_character_state: node["ratingCharacterState"]&.downcase,
          rating_overall_state: node["ratingOverallState"]&.downcase,
          body_html: node["bodyHtml"],
          likes_count: node["likesCount"],
          created_at: DateTime.parse(node["createdAt"]),
          modified_at: node["modifiedAt"] ? DateTime.parse(node["modifiedAt"]) : nil,
          viewer_did_like: node["viewerDidLike"],
          user: {
            username: node.dig("user", "username"),
            name: node.dig("user", "name"),
            avatar_url: node.dig("user", "avatarUrl"),
            is_supporter: node.dig("user", "isSupporter"),
          },
          record: {
            id: node.dig("record", "annictId"),
          }
        }
      end

      node["seriesList"]["nodes"].map do |node|
        work[:series_list] << {
          name: node["name"],
          name_en: node["nameEn"],
          series_works: node.dig("works", "edges").map do |work_edge|
            {
              summary: work_edge["summary"],
              summary_en: work_edge["summaryEn"],
              id: work_edge.dig("node", "annictId"),
              title: work_edge.dig("node", "title"),
              title_en: work_edge.dig("node", "titleEn"),
              image_url: work_edge.dig("node", "image", "internalUrl"),
            }
          end
        }
      end

      WorkEntity.new(work)
    end

    private

    def query
      load_query "work_detail/fetch_work.graphql"
    end
  end
end
