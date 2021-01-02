# frozen_string_literal: true

module V4
  class RecordsController < V4::ApplicationController
    def index
      set_page_category PageCategory::RECORD_LIST

      user = User.only_kept.find_by!(username: params[:username])

      @months = user.records.only_kept.group_by_month(:created_at, time_zone: user.time_zone).count.to_a.reverse.to_h

      result = Rails.cache.fetch(user_cache_key(user), expires_in: 3.hours) do
        RecordListPage::UserRepository.new(graphql_client: graphql_client).execute(username: user.username)
      end
      @user_entity = result.user_entity

      result = RecordListPage::RecordsRepository.
        new(graphql_client: graphql_client).
        execute(
          username: user.username,
          pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30),
          month: params[:month]
        )
      @record_entities = result.record_entities
      @page_info_entity = result.page_info_entity
    end

    def show
      set_page_category PageCategory::RECORD

      user = User.only_kept.find_by!(username: params[:username])
      record = user.records.only_kept.find(params[:record_id])

      @months = user.records.only_kept.group_by_month(:created_at, time_zone: user.time_zone).count.to_a.reverse.to_h

      result = RecordPage::UserRepository.new(graphql_client: graphql_client).execute(username: user.username)
      @user_entity = result.user_entity

      result = RecordPage::RecordRepository.
        new(graphql_client: graphql_client).
        execute(username: user.username, record_database_id: record.id)
      @record_entity = result.record_entity
    end

    private

    def user_cache_key(user)
      [
        PageCategory::RECORD_LIST,
        "user",
        user.id,
        user.updated_at.rfc3339
      ].freeze
    end
  end
end
