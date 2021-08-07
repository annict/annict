# frozen_string_literal: true

module GraphqlResolvable
  extend ActiveSupport::Concern

  def global_id
    Canary::AnnictSchema.id_from_object(self, self.class)
  end
end
