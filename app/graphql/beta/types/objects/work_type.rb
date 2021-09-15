# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class WorkType < Beta::Types::Objects::Base
        description "An anime title"

        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :title, String, null: false
        field :title_kana, String, null: true
        field :title_ro, String, null: true
        field :title_en, String, null: true
        field :media, Beta::Types::Enums::Media, null: false
        field :season_year, Integer, null: true
        field :season_name, Beta::Types::Enums::SeasonName, null: true
        field :official_site_url, String, null: true
        field :official_site_url_en, String, null: true
        field :wikipedia_url, String, null: true
        field :wikipedia_url_en, String, null: true
        field :twitter_username, String, null: true
        field :twitter_hashtag, String, null: true
        field :syobocal_tid, Integer, null: true
        field :mal_anime_id, String, null: true
        field :image, Beta::Types::Objects::WorkImageType, null: true
        field :satisfaction_rate, Float, null: true
        field :episodes_count, Integer, null: false
        field :watchers_count, Integer, null: false
        field :reviews_count, Integer, null: false
        field :no_episodes, Boolean, null: false
        field :viewer_status_state, Beta::Types::Enums::StatusState, null: true

        field :episodes, Beta::Types::Objects::EpisodeType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::EpisodeOrder, required: false
        end

        field :reviews, Beta::Types::Objects::ReviewType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::ReviewOrder, required: false
          argument :has_body, Boolean, required: false
        end

        field :programs, Beta::Types::Objects::ProgramType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::ProgramOrder, required: false
        end

        field :casts, Beta::Types::Objects::CastType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::CastOrder, required: false
        end

        field :staffs, Beta::Types::Objects::StaffType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::StaffOrder, required: false
        end

        field :series_list, Beta::Types::Objects::SeriesType.connection_type, null: true

        def episodes(order_by: nil)
          SearchEpisodesQuery.new(object.episodes, order_by: order_by).call
        end

        def reviews(order_by: nil, has_body: nil)
          SearchWorkRecordsQuery.new(WorkRecord.merge(object.records), order_by: order_by, has_body: has_body).call
        end

        def programs(order_by: nil)
          SlotsQuery.new(
            object.slots.only_kept,
            order: build_order(order_by)
          ).call
        end

        def casts(order_by: nil)
          SearchCastsQuery.new(object.casts, order_by: order_by).call
        end

        def staffs(order_by: nil)
          SearchStaffsQuery.new(object.staffs, order_by: order_by).call
        end

        def series_list
          object.series_list.only_kept.where("series_works_count > ?", 1)
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
          Beta::RecordLoader.for(WorkImage).load(object.work_image&.id)
        end

        def reviews_count
          object.work_records_count
        end

        def no_episodes
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
