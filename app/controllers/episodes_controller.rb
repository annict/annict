# frozen_string_literal: true

class EpisodesController < ApplicationV6Controller
  include EpisodeRecordListSettable
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::EPISODE_LIST

    set_work_header_resources
    raise ActionController::RoutingError, "Not Found" if @work.no_episodes?

    @work_ids = [@work.id]
    @episodes = @work.episodes.only_kept.order(:sort_number).page(params[:page]).per(100)
  end

  def show
    set_page_category PageCategory::EPISODE

    @work = Work.only_kept.find(params[:work_id])
    @work_ids = [@work.id]
    @programs = @work.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))

    @episode = @work.episodes.only_kept.find(params[:episode_id])
    @form = Forms::EpisodeRecordForm.new(episode: @episode, share_to_twitter: current_user&.share_record_to_twitter?)

    set_episode_record_list(@episode)
  end
end
