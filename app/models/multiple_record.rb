# == Schema Information
#
# Table name: multiple_records
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_multiple_records_on_user_id  (user_id)
#

class MultipleRecord < ActiveRecord::Base
end
