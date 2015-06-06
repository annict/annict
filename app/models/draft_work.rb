# == Schema Information
#
# Table name: draft_works
#
#  id                :integer          not null, primary key
#  work_id           :integer
#  season_id         :integer
#  sc_tid            :integer
#  title             :string           not null
#  media             :integer          not null
#  official_site_url :string           default(""), not null
#  wikipedia_url     :string           default(""), not null
#  released_at       :date
#  twitter_username  :string
#  twitter_hashtag   :string
#  released_at_about :string
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_draft_works_on_sc_tid     (sc_tid) UNIQUE
#  index_draft_works_on_season_id  (season_id)
#  index_draft_works_on_work_id    (work_id)
#

class DraftWork < ActiveRecord::Base
  include WorkCommon

  belongs_to :origin, foreign_key: :work_id, foreign_type: "Work"
  has_one :edit_request, as: :draft_resource

  accepts_nested_attributes_for :edit_request

  def to_diffable_hash
    self.class::DIFF_FIELDS.inject({}) do |hash, field|
      hash[field] = case field
      when :media
        send(field).to_s
      when :released_at
        send(field).try(:strftime, "%Y/%m/%d")
      else
        send(field)
      end

      hash
    end
  end
  
  def translated_diff_fields
    binding.pry
  end
end
