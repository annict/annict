# typed: false
# frozen_string_literal: true

describe "POST /graphql", type: :request do
  let(:work) { create(:work) }
  let(:access_token) { create(:oauth_access_token) }
  let(:id) { Beta::AnnictSchema.id_from_object(work, work.class) }
  let(:query) do
    <<~GRAPHQL
      query($workId: ID!) {
        node(id: $workId) {
          id
          ... on Work {
            annictId
            title
          }
        }
      }
    GRAPHQL
  end
  let(:headers) { {"Authorization" => "bearer #{access_token.token}"} }

  it "リクエストできること" do
    post "/graphql", params: {variables: {workId: id}, query: query}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to include({
      data: {
        node: {
          annictId: work.id,
          title: work.title,
          id: id
        }
      }
    }.deep_stringify_keys)
  end
end
