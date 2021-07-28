# frozen_string_literal: true

module AnimeHeaderLoadable
  extend ActiveSupport::Concern

  private

  def set_anime_header_resources
    @anime = Anime.only_kept.find(params[:anime_id])
    @programs = @anime.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))
  end
end
