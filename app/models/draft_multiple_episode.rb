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
  include DraftCommon
  include MultipleEpisodesFormatter

  DIFF_FIELDS = %i(body)

  belongs_to :work

  validates :body, presence: true, multiple_episode: true

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.inject({}) do |hash, field|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
