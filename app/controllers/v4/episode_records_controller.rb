# frozen_string_literal: true

module V4
  class EpisodeRecordsController < V4::ApplicationController
    include AnimeSidebarDisplayable
    include EpisodeDisplayable

    before_action :authenticate_user!, only: %i(create edit update switch)

    def create
      set_page_category PageCategory::EPISODE

      @form = EpisodeRecordForm.new(episode_record_form_attributes)

      unless @form.valid?
        load_episode_and_records(work_id: params[:work_id], episode_id: params[:id], form: @form)

        return render "/v4/episodes/show"
      end

      episode_record, err = CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: current_user)
      ).execute(form: @form)

      if err
        @form.error_messages = [err.message]
        load_episode_and_records(work_id: params[:work_id], episode_id: params[:id], form: @form)

        return render "/v4/episodes/show"
      end

      flash[:notice] = t("messages.episode_records.created")
      redirect_to episode_path(params[:work_id], params[:id])
    end

    private

    def episode_record_form_attributes
      @episode_record_form_attributes ||= params.to_unsafe_h["episode_record_form"]
    end
  end
end
