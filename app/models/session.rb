# == Schema Information
#
# Table name: sessions
#
#  id         :bigint(8)        not null, primary key
#  session_id :string           not null
#  data       :jsonb            not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_sessions_on_session_id  (session_id) UNIQUE
#  index_sessions_on_updated_at  (updated_at)
#

class Session < ApplicationRecord
end
