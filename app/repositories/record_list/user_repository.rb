# frozen_string_literal: true

module RecordList
  class UserRepository < ApplicationRepository
    def execute(username:)
      result = query(variables: { username: username })
      user_node = result.to_h.dig("data", "user")

      UserEntity.from_node(user_node)
    end
  end
end
