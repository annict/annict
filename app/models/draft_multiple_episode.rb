require "csv"

# == Schema Information
#
# Table name: draft_multiple_episodes
#
#  id         :integer          not null, primary key
#  work_id    :integer          not null
#  body       :text             not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_draft_multiple_episodes_on_work_id  (work_id)
#

class DraftMultipleEpisode < ActiveRecord::Base
  DIFF_FIELDS = %i(body)

  belongs_to :work
  has_one :edit_request, as: :draft_resource

  accepts_nested_attributes_for :edit_request

  validates :body, presence: true, multiple_episode: true

  def to_episode_hash
    body = self.body.gsub(/([^\\])\"/, %q/\\1__double_quote__/)

    CSV.parse(body).map do |ary|
      title = ary[1].gsub("__double_quote__", '"').try(:strip) if ary[1].present?
      { number: ary[0].try(:strip), title: title }
    end
  end

  def to_diffable_hash
    self.class::DIFF_FIELDS.inject({}) do |hash, field|
      hash[field] = send(field)
      hash
    end
  end
end
