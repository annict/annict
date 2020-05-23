# frozen_string_literal: true

module WorkDetail
  class FetchWorkRepository < ApplicationRepository
    def fetch(work_id:)
      result = execute(variables: { annictId: work_id.to_i })
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
        started_on: node["startedOn"],
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
        image_url_1x: node.dig("image", "internalUrl1x"),
        image_url_2x: node.dig("image", "internalUrl2x"),
        trailers: [],
        casts: [],
        staffs: [],
        episodes: [],
        programs: [],
        work_records: [],
        series_list: []
      }

      node["trailers"]["nodes"].map do |child_node|
        work[:trailers] << {
          title: child_node["title"],
          url: child_node["url"],
          image_url: child_node["internalImageUrl"]
        }
      end

      node["casts"]["nodes"].map do |child_node|
        work[:casts] << {
          accurate_name: child_node["accurateName"],
          accurate_name_en: child_node["accurateNameEn"],
          character: {
            id: child_node.dig("character", "annictId"),
            name: child_node.dig("character", "name"),
            name_en: child_node.dig("character", "nameEn")
          },
          person: {
            id: child_node.dig("person", "annictId")
          }
        }
      end

      node["staffs"]["nodes"].map do |child_node|
        work[:staffs] << {
          accurate_name: child_node["accurateName"],
          accurate_name_en: child_node["accurateNameEn"],
          role: child_node["role"],
          role_en: child_node["roleEn"],
          resource: {
            typename: child_node.dig("resource", "__typename"),
            id: child_node.dig("resource", "annictId")
          }
        }
      end

      node["episodes"]["nodes"].map do |child_node|
        work[:episodes] << {
          id: child_node["annictId"],
          number_text: child_node["numberText"],
          title: child_node["title"],
          title_en: child_node["titleEn"]
        }
      end

      node["programs"]["nodes"].map do |child_node|
        work[:programs] << {
          vod_title_name: child_node["vodTitleName"],
          vod_title_url: child_node["vodTitleUrl"],
          channel: {
            id: child_node.dig("channel", "annictId"),
            name: child_node.dig("channel", "name")
          }
        }
      end

      node["workRecords"]["nodes"].map do |child_node|
        work[:work_records] << {
          id: child_node["annictId"],
          rating_animation_state: child_node["ratingAnimationState"]&.downcase,
          rating_music_state: child_node["ratingMusicState"]&.downcase,
          rating_story_state: child_node["ratingStoryState"]&.downcase,
          rating_character_state: child_node["ratingCharacterState"]&.downcase,
          rating_overall_state: child_node["ratingOverallState"]&.downcase,
          body_html: child_node["bodyHtml"],
          likes_count: child_node["likesCount"],
          created_at: child_node["createdAt"],
          modified_at: child_node["modifiedAt"],
          user: {
            username: child_node.dig("user", "username"),
            name: child_node.dig("user", "name"),
            avatar_url: child_node.dig("user", "avatarUrl"),
            display_supporter_badge: child_node.dig("user", "displaySupporterBadge")
          },
          record: {
            id: child_node.dig("record", "annictId")
          }
        }
      end

      node["seriesList"]["nodes"].map do |child_node|
        work[:series_list] << {
          name: child_node["name"],
          name_en: child_node["nameEn"],
          series_works: child_node.dig("works", "edges").map do |work_edge|
            {
              summary: work_edge["summary"],
              summary_en: work_edge["summaryEn"],
              id: work_edge.dig("node", "annictId"),
              title: work_edge.dig("node", "title"),
              title_en: work_edge.dig("node", "titleEn"),
              image_url_1x: work_edge.dig("node", "image", "internalUrl1x"),
              image_url_2x: work_edge.dig("node", "image", "internalUrl2x"),
            }
          end
        }
      end

      WorkEntity.new(work)
    end
  end
end
