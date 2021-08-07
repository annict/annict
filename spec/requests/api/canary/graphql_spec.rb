# frozen_string_literal: true

describe "POST /canary/graphql", type: :request do
  let(:anime) { create(:anime) }
  let(:access_token) { create(:oauth_access_token) }
  let(:id) { Canary::AnnictSchema.id_from_object(anime, anime.class) }
  let(:query) do
    <<~GRAPHQL
      query($animeId: ID!) {
        node(id: $animeId) {
          ... on Anime {
            id
            databaseId
            title
          }
        }
      }
    GRAPHQL
  end
  let(:headers) { {"Authorization" => "bearer #{access_token.token}"} }

  it "リクエストできること" do
    post "/canary/graphql", params: {variables: {animeId: id}, query: query}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to include({
      data: {
        node: {
          databaseId: anime.id,
          title: anime.title,
          id: id
        }
      }
    }.deep_stringify_keys)
  end
end
