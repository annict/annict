# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class WorkType < Canary::Types::Objects::Base
        description "作品情報"

        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer,
          null: false,
          description: "Annict ID"

        field :title, String,
          null: false,
          description: "タイトル"

        field :title_Kana, String,
          null: true,
          description: "タイトル (かな)"

        field :title_en, String, null: true,
          description: "タイトル (英語)"

        field :title_ro, String,
          null: true,
          deprecation_reason: "このフィールドは使われていません。 `titleEn` を使用してください。"

        field :media, Canary::Types::Enums::Media,
          null: false

        field :season_year, Integer,
          null: true

        field :season_name, Canary::Types::Enums::SeasonName,
          null: true

        field :started_on, Canary::Types::Scalars::Date,
          null: true

        field :official_site_url, String,
          null: true

        field :official_site_url_en, String,
          null: true

        field :wikipedia_url, String,
          null: true

        field :wikipedia_url_en, String,
          null: true

        field :twitter_username, String,
          null: true

        field :twitter_hashtag, String,
          null: true

        field :syobocal_tid, Integer,
          null: true

        field :mal_anime_id, String,
          null: true

        field :image, Canary::Types::Objects::WorkImageType,
          null: true

        field :copyright, String,
          null: true

        field :satisfaction_rate, Float,
          null: true

        field :ratings_count, Integer,
          null: false,
          description: "評価数"

        field :episodes_count, Integer,
          null: false

        field :watchers_count, Integer,
          null: false

        field :work_records_count, Integer,
          null: false

        field :is_no_episodes, Boolean,
          null: false

        field :viewer_status_state, Canary::Types::Enums::StatusState,
          null: true

        field :synopsis, String,
          null: false

        field :synopsis_en, String,
          null: false

        field :synopsis_source, String,
          null: false

        field :synopsis_source_en, String,
          null: false

        field :episodes, Canary::Types::Objects::EpisodeType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeOrder, required: false
        end

        field :work_records, Canary::Types::Objects::WorkRecordType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::WorkRecordOrder, required: false
          argument :has_body, Boolean, required: false
        end

        field :programs, Canary::Types::Objects::ProgramType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ProgramOrder, required: false
        end

        field :casts, Canary::Types::Objects::CastType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::CastOrder, required: false
        end

        field :staffs, Canary::Types::Objects::StaffType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::StaffOrder, required: false
        end

        field :series_list, Canary::Types::Objects::SeriesType.connection_type, null: true

        field :trailers, Canary::Types::Objects::TrailerType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::TrailerOrder, required: false
        end

        def episodes(order_by: nil)
          SearchEpisodesQuery.new(object.episodes, order_by: order_by).call
        end

        def work_records(order_by: nil, has_body: nil)
          SearchWorkRecordsQuery.new(object.work_records, order_by: order_by, has_body: has_body).call
        end

        def programs(order_by: nil)
          SearchProgramsQuery.new(
            object.programs,
            order_by: order_by
          ).call
        end

        def casts(order_by: nil)
          SearchCastsQuery.new(object.casts, order_by: order_by).call
        end

        def staffs(order_by: nil)
          SearchStaffsQuery.new(object.staffs, order_by: order_by).call
        end

        def series_list
          object.series_list.published.where("series_works_count > ?", 1)
        end

        def trailers(order_by: nil)
          SearchTrailersQuery.new(object.pvs, order_by: order_by).call
        end

        def media
          object.media.upcase
        end

        def season_name
          object.season_name&.upcase
        end

        def syobocal_tid
          object.sc_tid
        end

        def image
          Canary::RecordLoader.for(WorkImage).load(object.work_image&.id)
        end

        def copyright
          object.work_image&.copyright
        end

        def is_no_episodes
          object.no_episodes?
        end

        def viewer_status_state
          state = context[:viewer].status_kind(object)
          state == "no_select" ? "NO_STATE" : state.upcase
        end
      end
    end
  end
end
