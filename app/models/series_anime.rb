# frozen_string_literal: true
# == Schema Information
#
# Table name: series_animes
#
#  id             :bigint           not null, primary key
#  aasm_state     :string           default("published"), not null
#  deleted_at     :datetime
#  summary        :string           default(""), not null
#  summary_en     :string           default(""), not null
#  unpublished_at :datetime
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  anime_id       :bigint           not null
#  series_id      :bigint           not null
#
# Indexes
#
#  index_series_animes_on_anime_id                (anime_id)
#  index_series_animes_on_deleted_at              (deleted_at)
#  index_series_animes_on_series_id               (series_id)
#  index_series_animes_on_series_id_and_anime_id  (series_id,anime_id) UNIQUE
#  index_series_animes_on_unpublished_at          (unpublished_at)
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#  fk_rails_...  (series_id => series.id)
#

class SeriesAnime < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i(anime_id).freeze

  counter_culture :series, column_name: -> (series_work) { series_work.published? ? :series_works_count : nil }

  belongs_to :series
  belongs_to :anime, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  def self.sort_season(sort_type: "ASC")
    joins(:anime).
      order("animes.season_year #{sort_type}").
      order("animes.season_name #{sort_type}")
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
