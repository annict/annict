# == Schema Information
#
# Table name: new_records
#
#  id                :integer          not null, primary key
#  user_id           :integer          not null
#  work_id           :integer          not null
#  aasm_state        :string           default("published"), not null
#  impressions_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_new_records_on_user_id  (user_id)
#  index_new_records_on_work_id  (work_id)
#

class NewRecord < ApplicationRecord
end
