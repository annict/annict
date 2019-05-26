# frozen_string_literal: true
# == Schema Information
#
# Table name: pvs
#
#  id                     :bigint(8)        not null, primary key
#  work_id                :integer          not null
#  url                    :string           not null
#  title                  :string           not null
#  thumbnail_file_name    :string
#  thumbnail_content_type :string
#  thumbnail_file_size    :integer
#  thumbnail_updated_at   :datetime
#  sort_number            :integer          default(0), not null
#  aasm_state             :string           default("published"), not null
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  title_en               :string           default(""), not null
#  image_data             :text
#
# Indexes
#
#  index_pvs_on_work_id  (work_id)
#

class Pv < ApplicationRecord
  include PvImageUploader::Attachment.new(:image)
  include AASM
  include DbActivityMethods
  include ImageUploadable

  DIFF_FIELDS = %i(title url sort_number).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  belongs_to :work

  validates :title, presence: true
  validates :url, url: true

  before_save :attach_thumbnail

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  def youtube?
    url.match?(%r{\A(https?\:\/\/)?(www\.youtube\.com|youtu\.?be)\/.+\z})
  end

  def youtube_video_id
    return unless youtube?
    params = Rack::Utils.parse_query(URI(url).query)
    params["v"]
  end

  private

  def attach_thumbnail
    return unless youtube?
    return unless attribute_changed?(:url)

    image_url = ""

    %w(
      maxresdefault
      sddefault
      hqdefault
      mqdefault
      default
    ).each do |size|
      image_url = "http://i.ytimg.com/vi/#{youtube_video_id}/#{size}.jpg"
      res = HTTParty.get(image_url)
      break if res.code == 200
    end

    self.thumbnail = URI.parse(image_url)
  end
end
