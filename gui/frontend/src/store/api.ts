import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { Session } from '../interfaces/interfaces';

const baseUrl = '//localhost:8080/';

export const api = createApi({
  baseQuery: fetchBaseQuery({ baseUrl }),
  tagTypes: ['Session', 'Phonon'],
  endpoints: (builder) => ({
    fetchSessions: builder.query<Session[], void>({
      query: () => 'listSessions',
      providesTags: ['Session'],
    }),
    unlockSession: builder.mutation<void, { sessionId: string; pin: string }>({
      query: ({ sessionId, pin }) => ({
        url: `cards/${sessionId}/unlock`,
        method: 'POST',
        body: { pin },
      }),
      invalidatesTags: ['Session'],
    }),
  }),
});

export const { useFetchSessionsQuery } = api;
