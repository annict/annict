# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "node" do
    let(:work) { create(:work) }
    let(:id) { GraphQL::Schema::UniqueWithinType.encode(work.class.name, work.id) }
    let(:result) do
      query_string = <<~GRAPHQL
        query {
          node(id: "#{id}") {
            id
            ... on Work {
              annictId
              title
            }
          }
        }
      GRAPHQL

      res = AnnictSchema.execute(query_string)
      pp(res) if res["errors"]
      res
    end

    it "fetches resource" do
      expected = {
        data: {
          node: {
            id: id,
            annictId: work.id,
            title: work.title
          }
        }
      }
      expect(result.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
    end
  end
end
