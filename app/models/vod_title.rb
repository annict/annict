# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: vod_titles
#
#  id             :bigint           not null, primary key
#  aasm_state     :string           default("published"), not null
#  code           :string           not null
#  deleted_at     :datetime
#  mail_sent_at   :datetime
#  name           :string           not null
#  unpublished_at :datetime
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#  channel_id     :bigint           not null
#  work_id        :bigint
#
# Indexes
#
#  index_vod_titles_on_channel_id      (channel_id)
#  index_vod_titles_on_deleted_at      (deleted_at)
#  index_vod_titles_on_mail_sent_at    (mail_sent_at)
#  index_vod_titles_on_unpublished_at  (unpublished_at)
#  index_vod_titles_on_work_id         (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (channel_id => channels.id)
#  fk_rails_...  (work_id => works.id)
#

class VodTitle < ApplicationRecord
  include SoftDeletable

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
