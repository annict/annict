# frozen_string_literal: true
# == Schema Information
#
# Table name: vod_titles
#
#  id         :integer          not null, primary key
#  channel_id :integer          not null
#  work_id    :integer
#  code       :string           not null
#  name       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_vod_titles_on_channel_id  (channel_id)
#  index_vod_titles_on_work_id     (work_id)
#

class VodTitle < ApplicationRecord
  belongs_to :channel
  belongs_to :work, optional: true
end
