# frozen_string_literal: true
# == Schema Information
#
# Table name: vod_titles
#
#  id           :bigint(8)        not null, primary key
#  channel_id   :bigint(8)        not null
#  work_id      :bigint(8)
#  code         :string           not null
#  name         :string           not null
#  aasm_state   :string           default("published"), not null
#  mail_sent_at :datetime
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#
# Indexes
#
#  index_vod_titles_on_channel_id    (channel_id)
#  index_vod_titles_on_mail_sent_at  (mail_sent_at)
#  index_vod_titles_on_work_id       (work_id)
#

class VodTitle < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :channel
  belongs_to :work, optional: true

  def import_csv
    [
      channel_id,
      nil,
      nil,
      code,
      name
    ].join(",")
  end
end
