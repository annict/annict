# frozen_string_literal: true

module V4
  class RecordsController < V4::ApplicationController
    def index
      set_page_category Rails.configuration.page_categories.record_list

      user = User.only_kept.find_by!(username: params[:username])

      @months = user.records.only_kept.group_by_month(:created_at, time_zone: user.time_zone).count.to_a.reverse.to_h

      @user_entity = Rails.cache.fetch(user_cache_key(user), expires_in: 3.hours) do
        RecordList::FetchUserRepository.new(graphql_client: graphql_client).fetch(username: user.username)
      end

      @record_entities, @page_info_entity = RecordList::FetchRecordsRepository.
        new(graphql_client: graphql_client).
        fetch(
          username: user.username,
          pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30),
          month: params[:month]
        )
    end

    private

    def user_cache_key(user)
      [
        "record-list",
        "user",
        user.id,
        user.updated_at.rfc3339
      ].freeze
    end
  end
end
