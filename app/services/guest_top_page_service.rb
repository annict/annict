# frozen_string_literal: true

class GuestTopPageService
  def self.season_top_work
    Work.only_kept.
      by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).
      order(watchers_count: :desc).
      first
  end

  def self.season_works
    Work.only_kept.
      by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).
      where.not(id: season_top_work.id).
      order(watchers_count: :desc).
      limit(12)
  end

  def self.top_work
    Work.only_kept.order(watchers_count: :desc).first
  end

  def self.works
    Work.only_kept.
      where.not(id: top_work.id).
      order(watchers_count: :desc).
      limit(12)
  end

  def self.cover_image_work
    cover_image_work_ids = [season_top_work.id, season_works.pluck(:id)].flatten
    Work.find(cover_image_work_ids.sample)
  end
end
