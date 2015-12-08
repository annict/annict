module Annict
  module Keen
    class Client
      def initialize(request)
        @request = request
      end

      def users
        @users ||= ::Annict::Keen::Collections::UsersCollection.new(@request)
      end

      def records
        @records ||= ::Annict::Keen::Collections::RecordsCollection.new(@request)
      end

      def follows
        @follows ||= ::Annict::Keen::Collections::FollowsCollection.new(@request)
      end

      def likes
        @likes ||= ::Annict::Keen::Collections::LikesCollection.new(@request)
      end

      def statuses
        @statuses ||= ::Annict::Keen::Collections::StatusesCollection.new(@request)
      end
    end
  end
end
