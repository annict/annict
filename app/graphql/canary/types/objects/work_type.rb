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

        field :title_en, String,
          null: true,
          description: "タイトル (英語)"

        field :local_title, String,
          null: false

        field :title_ro, String,
          null: true,
          deprecation_reason: "このフィールドは使われていません。 `titleEn` を使用してください。"

        field :media, Canary::Types::Enums::Media,
          null: false

        field :season_year, Integer,
          null: true

        field :season_name, Canary::Types::Enums::SeasonName,
          null: true

        field :local_season_name, String,
          null: true

        field :season_slug, String,
          null: true

        field :local_started_on_label, String,
          null: false,
          description: "開始日を表示する際のラベル名。テレビの場合は「放送開始日」、映画の場合は「公開日」といった感じでメディアごとに少し文言が変わる"

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

        field :work_records_with_body_count, Integer,
          null: false

        field :is_no_episodes, Boolean,
          null: false

        field :viewer_finished_to_watch, Boolean,
          null: false

        field :viewer_status_kind, Canary::Types::Enums::StatusKind,
          null: true

        field :synopsis, String,
          null: false

        field :synopsis_en, String,
          null: false

        field :local_synopsis, String,
          null: false

        field :local_synopsis_html, String,
          null: false

        field :synopsis_source, String,
          null: false

        field :synopsis_source_en, String,
          null: false

        field :local_synopsis_source, String,
          null: false

        field :episodes, Canary::Types::Objects::EpisodeType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeOrder, required: false
        end

        field :work_records, Canary::Types::Objects::WorkRecordType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::WorkRecordOrder, required: false
          argument :has_body, Boolean, required: false
          argument :filter_by_locale, Boolean, required: false
        end

        field :programs, Canary::Types::Objects::ProgramType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::ProgramOrder, required: false
        end

        field :slots, Canary::Types::Objects::SlotType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::SlotOrder, required: false
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

        def work_records(order_by: nil, has_body: nil, filter_by_locale: false)
          SearchWorkRecordsQuery.new(object.work_records,
            context: context,
            order_by: order_by,
            has_body: has_body,
            filter_by_locale: filter_by_locale
          ).call
        end

        def programs(order_by: nil)
          SearchProgramsRepository.new(
            object.program_details,
            order_by: order_by
          ).call
        end

        def slots(order_by: nil)
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

        def local_season_name
          object.season&.local_name
        end

        def season_slug
          object.season&.slug
        end

        def local_started_on_label
          object.decorate.started_on_label
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

        def viewer_finished_to_watch
          return false unless context[:viewer]

          context[:viewer].latest_statuses.finished_to_watch.where(work_id: object.id).exists? &&
            context[:viewer].work_records.published.where(work_id: object.id).exists?
        end

        def viewer_status_kind
          return unless context[:viewer]

          kind = context[:viewer].status_kind_v3(object)
          kind == "no_status" ? "NO_STATUS" : kind.upcase
        end

        def local_synopsis
          object.local_synopsis(raw: true)
        end

        def local_synopsis_html
          object.local_synopsis(raw: false)
        end
      end
    end
  end
end
