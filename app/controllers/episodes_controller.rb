# frozen_string_literal: true

class EpisodesController < V4::ApplicationController
  def show
    set_page_category PageCategory::EPISODE

    @anime = Anime.only_kept.find(params[:anime_id])
    @episode = @anime.episodes.only_kept.find(params[:episode_id])
    @vod_channels = Channel.only_kept.eager_load(:programs).merge(@anime.programs.only_kept.in_vod).order(:sort_number)
    @public_records = @episode.records.only_kept.eager_load(:episode_record, user: %i(gumroad_subscriber profile setting)).
      merge(EpisodeRecord.with_body.order_by_rating_state(:desc).order(created_at: :desc)).
      limit(30)
  end
end
