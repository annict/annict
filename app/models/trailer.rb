# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: trailers
#
#  id                     :bigint           not null, primary key
#  aasm_state             :string           default("published"), not null
#  deleted_at             :datetime
#  image_data             :text
#  sort_number            :integer          default(0), not null
#  thumbnail_content_type :string
#  thumbnail_file_name    :string
#  thumbnail_file_size    :integer
#  thumbnail_updated_at   :datetime
#  title                  :string           not null
#  title_en               :string           default(""), not null
#  unpublished_at         :datetime
#  url                    :string           not null
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  work_id                :bigint           not null
#
# Indexes
#
#  index_trailers_on_deleted_at      (deleted_at)
#  index_trailers_on_unpublished_at  (unpublished_at)
#  index_trailers_on_work_id         (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (work_id => works.id)
#

class Trailer < ApplicationRecord
  T.unsafe(self).include TrailerImageUploader::Attachment.new(:image)
  include DbActivityMethods
  include ImageUploadable
  include Unpublishable

  DIFF_FIELDS = %i[title url sort_number].freeze

  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  belongs_to :work, touch: true

  validates :title, presence: true
  validates :url, url: true

  before_save :attach_thumbnail

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def youtube?
    url.match?(%r{\A(https?://)?(www\.youtube\.com|youtu\.?be)/.+\z})
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

    %w[
      maxresdefault
      sddefault
      hqdefault
      mqdefault
      default
    ].each do |size|
      image_url = "http://i.ytimg.com/vi/#{youtube_video_id}/#{size}.jpg"
      res = HTTParty.get(image_url)
      break if res.code == 200
    end

    self.image = Down.open(image_url)
  end
end
