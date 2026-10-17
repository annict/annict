# typed: false
# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "node" do
    let(:work) { create(:work) }
    let(:id) { Beta::AnnictSchema.id_from_object(work, work.class) }
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

      res = Beta::AnnictSchema.execute(query_string)
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

  describe "node (Person)" do
    def execute_node_query(person)
      id = Beta::AnnictSchema.id_from_object(person, person.class)
      query_string = <<~GRAPHQL
        query {
          node(id: "#{id}") {
            ... on Person {
              name
            }
          }
        }
      GRAPHQL

      Beta::AnnictSchema.execute(query_string)
    end

    it "公開中の人物を返すこと" do
      person = create(:person)

      expect(execute_node_query(person).dig("data", "node", "name")).to eq(person.name)
    end

    it "非公開・削除済みの人物は、存在しない人物と同じく見つからない扱いになること" do
      unpublished_person = create(:person, :unpublished)
      deleted_person = create(:person, :deleted)

      expect { execute_node_query(unpublished_person) }.to raise_error(ActiveRecord::RecordNotFound)
      expect { execute_node_query(deleted_person) }.to raise_error(ActiveRecord::RecordNotFound)
    end
  end
end
