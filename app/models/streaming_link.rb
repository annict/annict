# frozen_string_literal: true
# == Schema Information
#
# Table name: streaming_links
#
#  id         :integer          not null, primary key
#  channel_id :integer          not null
#  work_id    :integer          not null
#  locale     :string           not null
#  unique_id  :string           not null
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_streaming_links_on_channel_id                           (channel_id)
#  index_streaming_links_on_channel_id_and_locale_and_unique_id  (channel_id,locale,unique_id) UNIQUE
#  index_streaming_links_on_channel_id_and_work_id_and_locale    (channel_id,work_id,locale) UNIQUE
#  index_streaming_links_on_work_id                              (work_id)
#

class StreamingLink < ApplicationRecord
  extend Enumerize
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(channel_id locale unique_id).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  enumerize :locale, in: %w(ja en), default: :ja

  validates :channel_id, presence: true
  validates :locale, presence: true, locale: true
  validates :unique_id, presence: true
  validates :work_id, presence: true

  belongs_to :channel
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  def url
    case channel_id
    when 107 then "http://www.b-ch.com/ttl/index.php?ttl_c=#{unique_id}"
    when 165 then "http://ch.nicovideo.jp/#{unique_id}"
    when 241 then "https://anime.dmkt-sp.jp/animestore/ci_pc?workId=#{unique_id}"
    when 243 then "https://www.amazon.co.jp/dp/#{unique_id}/"
    when 244 then "http://www.crunchyroll.com/#{unique_id}"
    when 245 then "http://www.daisuki.net/jp/en/anime/detail.#{unique_id}.html"
    when 246 then "https://www.funimation.com/#{unique_id}"
    when 247 then "http://www.hulu.jp/#{unique_id}"
    when 248 then "https://www.netflix.com/title/#{unique_id}"
    when 249 then "http://video.unext.jp/title/#{unique_id}"
    end
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
