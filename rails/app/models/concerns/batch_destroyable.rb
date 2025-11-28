# typed: false
# frozen_string_literal: true

module BatchDestroyable
  extend ActiveSupport::Concern

  included do
    def destroy_in_batches
      dependent_associations_to_destroy.each do |assoc|
        public_send(assoc.name).find_each(&:destroy_in_batches)
      end

      destroy
    end

    private

    def dependent_associations_to_destroy
      self.class.reflect_on_all_associations(:has_many).select { |assoc| assoc.options[:dependent] == :destroy }
    end
  end
end
