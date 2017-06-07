# frozen_string_literal: true
# == Schema Information
#
# Table name: pvs
#
#  id                     :integer          not null, primary key
#  work_id                :integer          not null
#  title                  :string           not null
#  url                    :string           not null
#  thumbnail_file_name    :string
#  thumbnail_content_type :string
#  thumbnail_file_size    :integer
#  thumbnail_updated_at   :datetime
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  sort_number            :integer          default(0), not null
#  aasm_state             :string           default("published"), not null
#
# Indexes
#
#  index_pvs_on_work_id  (work_id)
#

class Pv < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(title url sort_number).freeze

  has_attached_file :thumbnail

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
  validates :thumbnail, attachment_content_type: { content_type: /\Aimage/ }

  before_save :attach_thumbnail

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  private

  def attach_thumbnail
    return unless url.match?(%r{\A(https?\:\/\/)?(www\.youtube\.com|youtu\.?be)\/.+\z})

    params = Rack::Utils.parse_query(URI(url).query)
    image_url = "http://i.ytimg.com/vi/#{params['v']}/maxresdefault.jpg"

    self.thumbnail = URI.parse(image_url)
  end
end
