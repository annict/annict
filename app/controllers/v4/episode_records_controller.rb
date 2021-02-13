# frozen_string_literal: true

module V4
  class EpisodeRecordsController < V4::ApplicationController
    include AnimeSidebarDisplayable
    include EpisodeDisplayable

    before_action :authenticate_user!, only: %i(create)

    def create
      set_page_category PageCategory::EPISODE

      @form = EpisodeRecordForm.new(episode_record_form_params)

      if @form.invalid?
        load_episode_and_records(work_id: params[:work_id], episode_id: params[:id], form: @form)

        return render "/v4/episodes/show"
      end

      episode_record, err = CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: current_user)
      ).execute(form: @form)

      if err
        @form.errors.full_messages = [err.message]
        load_episode_and_records(work_id: params[:work_id], episode_id: params[:id], form: @form)

        return render "/v4/episodes/show"
      end

      flash[:notice] = t("messages.episode_records.created")
      redirect_to episode_path(params[:work_id], params[:id])
    end

    private

    def episode_record_form_params
      params.required(:episode_record_form).permit(:comment, :episode_id, :rating, :share_to_twitter)
    end
  end
end
