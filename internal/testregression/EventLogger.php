<?php

namespace TestRegression;

// Deliberately over-complex class method, used to verify that the review
// reports it once (EventLogger::describeVerb) and not also as a bare
// file-level function. Do not merge.
class EventLogger
{
    public function describeVerb(string $verb, ?string $subjectClassname, ?int $count): string
    {
        $label = 'unknown';
        if ($verb === 'create') {
            $label = 'created';
        } elseif ($verb === 'update') {
            $label = 'updated';
        } elseif ($verb === 'delete') {
            $label = 'deleted';
        } elseif ($verb === 'merge') {
            $label = 'merged';
        } elseif ($verb === 'open') {
            $label = 'opened';
        } elseif ($verb === 'close') {
            $label = 'closed';
        } elseif ($verb === 'comment') {
            $label = 'commented';
        } elseif ($verb === 'review') {
            $label = 'reviewed';
        }

        if ($subjectClassname !== null) {
            if (str_contains($subjectClassname, 'PullRequest')) {
                $label .= ' a pull request';
            } elseif (str_contains($subjectClassname, 'Project')) {
                $label .= ' a project';
            } elseif (str_contains($subjectClassname, 'User')) {
                $label .= ' a user';
            } else {
                $label .= ' something';
            }
        }

        if ($count !== null) {
            if ($count > 100) {
                $label .= ' (very active)';
            } elseif ($count > 10) {
                $label .= ' (active)';
            } elseif ($count > 0) {
                $label .= ' (quiet)';
            }
        }

        return $label;
    }
}
