# frozen_string_literal: true

module AnimeListable
  extend ActiveSupport::Concern

  included do
    before_action :set_display_option
  end

  private

  def set_display_option
    @display_option = params[:display].in?(%w[grid grid_small]) ? params[:display] : "grid"
  end

  def display_works_count
    @display_option == "grid" ? 30 : 120
  end

  def set_resource_data(animes)
    @anime_ids = animes.map(&:id)

    if @display_option == "grid"
      @casts_data = Cast.only_kept.where(anime: animes).order(:sort_number).group_by(&:work_id)
      @staffs_data = Staff.only_kept.major.where(anime: animes).order(:sort_number).group_by(&:work_id)
    end
  end
end
