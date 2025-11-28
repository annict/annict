# typed: false
# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "viewer" do
    let(:user) { create(:user) }
    let(:result) do
      query_string = <<~GRAPHQL
        query {
          viewer {
            annictId
            username
          }
        }
      GRAPHQL

      res = Beta::AnnictSchema.execute(query_string, context: {viewer: user})
      pp(res) if res["errors"]
      res
    end

    it "fetches user" do
      expected = {
        data: {
          viewer: {
            annictId: user.id,
            username: user.username
          }
        }
      }
      expect(result.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end
end
